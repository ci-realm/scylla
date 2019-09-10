module Page.Organizations exposing (view)

import Data exposing (Organization)
import Element
    exposing
        ( Element
        , alignRight
        , centerX
        , centerY
        , column
        , fill
        , maximum
        , padding
        , row
        , spacing
        , text
        , width
        )
import Element.Background as Background
import Init exposing (urlFor)
import Material.Icons.Av exposing (replay)
import Model exposing (Model)
import Msgs exposing (Msg(..), Route(..))
import Style exposing (iconButton, lightGrey, styledLink)


view : Model -> Element Msg
view model =
    column [ width (fill |> maximum 1000), centerX ]
        [ row [ width fill ] [ column [ width fill ] [ recentBuilds model ] ]
        , row [ width fill ] [ column [ alignRight ] [ iconButton replay "Refresh" <| Just GetOrganizations ] ]
        ]


recentBuilds : Model -> Element Msg
recentBuilds model =
    Element.table [ centerX, centerY, spacing 5, padding 10, Background.color lightGrey ]
        { data = model.organizations
        , columns =
            [ { header = text "Name", width = fill, view = orgName }
            , { header = text "URL", width = fill, view = orgURL }
            , { header = text "Builds", width = fill, view = orgBuildCount }
            ]
        }


orgName : Organization -> Element msg
orgName org =
    styledLink { url = urlFor <| OrganizationBuildsRoute org.owner, label = text org.owner }


orgURL : Organization -> Element msg
orgURL org =
    text org.url


orgBuildCount : Organization -> Element msg
orgBuildCount org =
    text <| String.fromInt org.buildCount
