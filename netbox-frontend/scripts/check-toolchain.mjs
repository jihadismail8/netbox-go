import process from 'node:process'
import packageMetadata from '../package.json' with { type: 'json' }

const expectedNode = `v${packageMetadata.engines.node}`
const npmAgent = process.env.npm_config_user_agent ?? ''
const actualNPM = npmAgent.match(/^npm\/([^ ]+)/)?.[1]
const failures = []

if (process.version !== expectedNode) {
  failures.push(
    `Node ${packageMetadata.engines.node} is required; found ${process.version.slice(1)}`,
  )
}
if (actualNPM !== packageMetadata.engines.npm) {
  failures.push(`npm ${packageMetadata.engines.npm} is required; found ${actualNPM ?? 'unknown'}`)
}

if (failures.length > 0) {
  for (const failure of failures) console.error(failure)
  process.exitCode = 1
} else {
  console.log(`Toolchain OK: Node ${process.version.slice(1)}, npm ${actualNPM}`)
}
