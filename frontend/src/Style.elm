module Style exposing (blue, fontSize, grey, iconButton, lightBlue, lightGrey, styledButton, styledLink, white)

import Element exposing (Element, el, link, modular, padding, rgb, row, spacing, text)
import Element.Background as Background
import Element.Font as Font
import Element.Input exposing (button)
import Html exposing (Html)
import Material.Icons exposing (Coloring(..))


fontSize : Int -> Element.Attr decorative msg
fontSize n =
    Font.size <| round <| modular 16 1.25 n


styledLink : { label : Element msg, url : String } -> Element msg
styledLink =
    link [ Font.color lightBlue ]


styledButton : { label : Element msg, onPress : Maybe msg } -> Element msg
styledButton =
    button [ padding 15, Font.color white, Background.color blue ]


iconButton : (Int -> Coloring -> Html msg) -> String -> Maybe msg -> Element msg
iconButton icon label onPress =
    styledButton
        { label = row [ spacing 10 ] [ el [] (Element.html <| icon 20 Inherit), el [] <| text label ]
        , onPress = onPress
        }


grey : Element.Color
grey =
    rgb 0.18 0.18 0.18


lightGrey : Element.Color
lightGrey =
    rgb 0.25 0.25 0.25


white : Element.Color
white =
    rgb 1 1 1


blue : Element.Color
blue =
    rgb 0.13 0.58 0.95


lightBlue : Element.Color
lightBlue =
    rgb 0.69 0.87 0.85
