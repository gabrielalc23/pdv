import { CardAction } from "./card-action"
import { CardContent } from "./card-content"
import { CardDescription } from "./card-description"
import { CardFooter } from "./card-footer"
import { CardHeader } from "./card-header"
import { CardRoot } from "./card-root"
import { CardTitle } from "./card-title"

type CardComponentType = typeof CardRoot & {
  Action: typeof CardAction
  Content: typeof CardContent
  Description: typeof CardDescription
  Footer: typeof CardFooter
  Header: typeof CardHeader
  Title: typeof CardTitle
}

export const Card: CardComponentType = Object.assign(CardRoot, {
  Action: CardAction,
  Content: CardContent,
  Description: CardDescription,
  Footer: CardFooter,
  Header: CardHeader,
  Title: CardTitle,
}) satisfies CardComponentType
