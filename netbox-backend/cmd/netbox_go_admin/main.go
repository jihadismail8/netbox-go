// Command netbox_go_admin performs local, one-time identity administration.
// It exposes no network listener and never accepts passwords in arguments.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/unix"

	"netbox-go/configs"
	identityapp "netbox-go/internal/application/identity"
	"netbox-go/internal/config"
	"netbox-go/internal/database"
	identitydomain "netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

const commandUsage = "usage: netbox_go_admin <bootstrap|reset-password|create-user|grant-permission> [options]"

type commandOptions struct {
	command       string
	configFile    string
	actorUsername string
	username      string
	email         string
	permission    string
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fail(err.Error())
	}
}

func run(arguments []string, stdin io.Reader, stdout, stderr io.Writer) error {
	options, err := parseCommand(arguments, stderr)
	if err != nil {
		return err
	}
	if err := config.Init(options.configFile); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if err := config.ApplyEnvironmentOverrides(); err != nil {
		return fmt.Errorf("load environment: %w", err)
	}
	// User creation persists a password hash. Keep SQL parameter logging off at
	// this administration boundary even when it is enabled for the application.
	config.Get().Database.Postgresql.EnableLog = false
	database.InitDB()
	defer func() { _ = database.CloseDB() }()
	if err := database.AutoMigrate(); err != nil {
		return fmt.Errorf("bootstrap database: %w", err)
	}

	core := composition.NewCore(database.GetDB())
	ctx := context.Background()
	reader := bufio.NewReader(stdin)
	switch options.command {
	case "bootstrap":
		password, err := readConfirmedPassword(reader, stdin, stderr, "Password: ", "Confirm password: ")
		if err != nil {
			return err
		}
		user, err := core.Identity.BootstrapAdministrator(ctx, options.username, options.email, password)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "administrator %q created (id %d)\n", user.Username, user.ID)
		return writeResultError(err)

	case "reset-password":
		password, err := readConfirmedPassword(reader, stdin, stderr, "Password: ", "Confirm password: ")
		if err != nil {
			return err
		}
		if err := core.Identity.ResetAdministratorPassword(ctx, options.username, password); err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "password reset for administrator %q\n", options.username)
		return writeResultError(err)

	case "create-user":
		actor, err := authenticateActor(ctx, core.Identity, reader, stdin, stderr, options.actorUsername)
		if err != nil {
			return err
		}
		password, err := readConfirmedPassword(reader, stdin, stderr, "New password: ", "Confirm new password: ")
		if err != nil {
			return err
		}
		user, err := core.Identity.CreateLocalUser(ctx, actor.Principal(), identityapp.CreateUserInput{
			Username: options.username,
			Email:    options.email,
			Password: password,
		})
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "user %q created (id %d)\n", user.Username, user.ID)
		return writeResultError(err)

	case "grant-permission":
		actor, err := authenticateActor(ctx, core.Identity, reader, stdin, stderr, options.actorUsername)
		if err != nil {
			return err
		}
		input, err := parsePermission(options.permission)
		if err != nil {
			return err
		}
		grant, err := core.Identity.GrantPermissionToUserByUsername(ctx, actor.Principal(), options.username, input)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "permission %q granted to user %q\n", grant.Codename(), options.username)
		return writeResultError(err)
	}
	return errors.New("unknown command: " + options.command)
}

func parseCommand(arguments []string, stderr io.Writer) (commandOptions, error) {
	if len(arguments) == 0 {
		return commandOptions{}, errors.New(commandUsage)
	}
	options := commandOptions{command: arguments[0], configFile: configs.Location("netbox_go.yml")}
	flags := flag.NewFlagSet(options.command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.configFile, "config", options.configFile, "configuration file")

	switch options.command {
	case "bootstrap":
		options.username = "admin"
		flags.StringVar(&options.username, "username", options.username, "administrator username")
		flags.StringVar(&options.email, "email", "", "administrator email")
	case "reset-password":
		options.username = "admin"
		flags.StringVar(&options.username, "username", options.username, "administrator username")
	case "create-user":
		flags.StringVar(&options.actorUsername, "actor-username", "", "existing administrator username")
		flags.StringVar(&options.username, "username", "", "new non-superuser username")
		flags.StringVar(&options.email, "email", "", "new non-superuser email")
	case "grant-permission":
		flags.StringVar(&options.actorUsername, "actor-username", "", "existing administrator username")
		flags.StringVar(&options.username, "username", "", "target username")
		flags.StringVar(&options.permission, "permission", "", "model permission (for example dcim.view_site)")
	default:
		return commandOptions{}, errors.New("unknown command: " + options.command)
	}

	if err := flags.Parse(arguments[1:]); err != nil {
		return commandOptions{}, err
	}
	if flags.NArg() != 0 {
		return commandOptions{}, errors.New("unexpected positional arguments; passwords must be provided on stdin")
	}
	if options.command == "create-user" || options.command == "grant-permission" {
		if strings.TrimSpace(options.actorUsername) == "" {
			return commandOptions{}, errors.New("--actor-username is required")
		}
		if strings.TrimSpace(options.username) == "" {
			return commandOptions{}, errors.New("--username is required")
		}
	}
	if options.command == "grant-permission" && strings.TrimSpace(options.permission) == "" {
		return commandOptions{}, errors.New("--permission is required")
	}
	return options, nil
}

func authenticateActor(ctx context.Context, service *identityapp.Service, reader *bufio.Reader, stdin io.Reader, stderr io.Writer, username string) (identitydomain.User, error) {
	password, err := readPassword(reader, stdin, stderr, "Administrator password: ")
	if err != nil {
		return identitydomain.User{}, err
	}
	user, err := service.AuthenticatePassword(ctx, username, password)
	if err != nil || !user.IsSuperuser {
		return identitydomain.User{}, errors.New("administrator authentication failed")
	}
	return user, nil
}

func parsePermission(value string) (identityapp.PermissionGrantInput, error) {
	value = strings.TrimSpace(value)
	appLabel, codename, found := strings.Cut(value, ".")
	if !found || strings.Contains(codename, ".") {
		return identityapp.PermissionGrantInput{}, errors.New("permission must use the form app_label.action_model")
	}
	action, model, found := strings.Cut(codename, "_")
	if !found || appLabel == "" || action == "" || model == "" {
		return identityapp.PermissionGrantInput{}, errors.New("permission must use the form app_label.action_model")
	}
	return identityapp.PermissionGrantInput{AppLabel: appLabel, Action: action, Model: model}, nil
}

func readConfirmedPassword(reader *bufio.Reader, stdin io.Reader, stderr io.Writer, passwordPrompt, confirmationPrompt string) (string, error) {
	first, err := readPassword(reader, stdin, stderr, passwordPrompt)
	if err != nil {
		return "", err
	}
	second, err := readPassword(reader, stdin, stderr, confirmationPrompt)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", errors.New("passwords do not match")
	}
	return first, nil
}

func readPassword(reader *bufio.Reader, stdin io.Reader, stderr io.Writer, prompt string) (string, error) {
	_, _ = fmt.Fprint(stderr, prompt)
	if file, ok := stdin.(*os.File); ok {
		fd := int(file.Fd())
		termios, termErr := unix.IoctlGetTermios(fd, unix.TCGETS)
		if termErr == nil {
			hidden := *termios
			hidden.Lflag &^= unix.ECHO
			if err := unix.IoctlSetTermios(fd, unix.TCSETS, &hidden); err != nil {
				return "", fmt.Errorf("protect password input: %w", err)
			}
			defer func() {
				_ = unix.IoctlSetTermios(fd, unix.TCSETS, termios)
				_, _ = fmt.Fprintln(stderr)
			}()
		}
	}
	value, err := reader.ReadString('\n')
	if err != nil && len(value) == 0 {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(value, "\r\n"), nil
}

func writeResultError(err error) error {
	if err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	return nil
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
