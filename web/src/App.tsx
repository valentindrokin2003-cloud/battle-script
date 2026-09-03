import { useState } from 'react'
import { BattleResultScreen } from './screens/BattleResultScreen'
import { BossSelect } from './screens/BossSelect'
import { IntentReview } from './screens/IntentReview'
import { TacticInput } from './screens/TacticInput'
import type { Boss, IntentByHero, TacticTexts } from './api/types'

type Screen =
  | { name: 'boss-select' }
  | { name: 'tactic-input'; boss: Boss }
  | { name: 'intent-review'; boss: Boss; texts: TacticTexts }
  | { name: 'battle-result'; boss: Boss; intents: IntentByHero }

export function App() {
  const [screen, setScreen] = useState<Screen>({ name: 'boss-select' })

  switch (screen.name) {
    case 'boss-select':
      return <BossSelect onSelect={(boss) => setScreen({ name: 'tactic-input', boss })} />

    case 'tactic-input':
      return (
        <TacticInput
          boss={screen.boss}
          onBack={() => setScreen({ name: 'boss-select' })}
          onSubmit={(texts) => setScreen({ name: 'intent-review', boss: screen.boss, texts })}
        />
      )

    case 'intent-review':
      return (
        <IntentReview
          boss={screen.boss}
          texts={screen.texts}
          onBack={() => setScreen({ name: 'tactic-input', boss: screen.boss })}
          onConfirm={(intents) => setScreen({ name: 'battle-result', boss: screen.boss, intents })}
        />
      )

    case 'battle-result':
      return (
        <BattleResultScreen
          boss={screen.boss}
          intents={screen.intents}
          onPlayAgain={() => setScreen({ name: 'boss-select' })}
        />
      )
  }
}
