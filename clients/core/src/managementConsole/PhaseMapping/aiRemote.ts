import { loadRemote, registerRemotes } from '@module-federation/runtime-tools/runtime'

let registered = false

// The federation runtime fetches the entry of every static remote on each page load, so the AI
// remote, which only exists while AI is on, is registered on first use instead.
export const loadAIRemote = async <T>(module: 'routes' | 'sidebar'): Promise<T> => {
  if (!registered) {
    registerRemotes([{ name: 'ai_component', entry: __PROMPT_AI_REMOTE_ENTRY__ }])
    registered = true
  }
  const loaded = await loadRemote<T>(`ai_component/${module}`)
  if (!loaded) throw new Error(`ai_component/${module} could not be loaded`)
  return loaded
}
