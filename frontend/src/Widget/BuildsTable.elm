module Widget.BuildsTable exposing (view)

import Data exposing (Build)
import Element exposing (Element, centerX, centerY, column, fill, padding, spacing, text)
import Element.Background as Background
import Init exposing (urlFor)
import Msgs exposing (Msg(..), Route(..))
import Style exposing (lightGrey, styledLink)
import Time
import Utils exposing (timeDistanceWithAffix, timeDistanceWithoutAffix)


view : { a | time : Time.Posix, builds : List Build } -> Element Msg
view model =
    Element.table [ centerX, centerY, spacing 5, padding 10, Background.color lightGrey ]
        { data = model.builds
        , columns =
            [ { header = text "Build", width = fill, view = recentBuildsBuildColumn }
            , { header = text "Status", width = fill, view = \build -> text build.status }
            , { header = text "Project", width = fill, view = recentBuildsProjectColumn }
            , { header = text "SHA", width = fill, view = recentBuildsSHAColumn }
            , { header = text "Started", width = fill, view = recentBuildsStartedAtColumn model }
            , { header = text "Duration", width = fill, view = recentBuildsDurationColumn }
            ]
        }



-- display pending if finishedAt is 0001-01-01 00:00:00 UTC


recentBuildsDurationColumn : Build -> Element Msg
recentBuildsDurationColumn build =
    if -62135596800000 == Time.posixToMillis build.finishedAt then
        text "pending"

    else
        text <| timeDistanceWithoutAffix build.createdAt build.finishedAt


recentBuildsStartedAtColumn : { a | time : Time.Posix } -> Build -> Element Msg
recentBuildsStartedAtColumn model build =
    text <| timeDistanceWithAffix model.time build.createdAt


recentBuildsSHAColumn : Build -> Element Msg
recentBuildsSHAColumn build =
    let
        shaLink =
            build.repoLink ++ "/commit/" ++ build.sha
    in
    styledLink { url = shaLink, label = text <| shortSHA build.sha }


shortSHA : String -> String
shortSHA =
    String.slice 0 7


recentBuildsBuildColumn : Build -> Element Msg
recentBuildsBuildColumn build =
    styledLink { url = urlFor <| BuildRoute build.id, label = text <| String.fromInt build.id }


recentBuildsProjectColumn : Build -> Element Msg
recentBuildsProjectColumn build =
    column []
        [ styledLink { url = build.ownerLink, label = text build.owner }
        , styledLink { url = build.repoLink, label = text build.repo }
        ]
