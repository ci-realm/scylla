module Page.Build exposing (view)

import Data exposing (Build, LogLine)
import Element
    exposing
        ( Element
        , centerX
        , centerY
        , column
        , fill
        , height
        , htmlAttribute
        , maximum
        , padding
        , paragraph
        , rgba
        , row
        , scrollbars
        , shrink
        , spacing
        , text
        , width
        )
import Element.Font as Font
import Html.Attributes exposing (id)
import Init exposing (urlFor)
import Model exposing (Model)
import Msgs exposing (Msg(..), Route(..))
import Style exposing (fontSize, styledButton, styledLink)
import Time
import Utils exposing (timeDistanceWithAffix, timeDistanceWithoutAffix)


view : Model -> Element Msg
view model =
    case model.build of
        Nothing ->
            column [ width fill ] [ text <| "Build not found" ]

        Just build ->
            column [ width fill ] (viewBuild model build)


viewBuild : Model -> Build -> List (Element Msg)
viewBuild model build =
    [ row []
        [ column [ fontSize 4, width fill ]
            [ paragraph []
                [ styledLink
                    { url = build.ownerLink
                    , label = text build.owner
                    }
                , text <| " / "
                , styledLink
                    { url = build.repoLink
                    , label = text build.repo
                    }
                , text <| " Build "
                , styledLink
                    { url = urlFor <| BuildRoute build.id
                    , label = text <| "#" ++ String.fromInt build.id
                    }
                ]
            ]
        ]
    , row [] [ column [] [ text <| "Pull Request " ++ String.fromInt build.prNumber, text <| build.prTitle ] ]
    , row [] [ column [] [ text <| "Commit " ++ build.sha ] ]
    , row [] [ column [] [ text <| "Branch " ++ build.branch ] ]
    , row [] [ column [] [ text <| build.userLogin ] ]
    , row [] [ column [] [ text <| "Finished in " ++ timeDistanceWithoutAffix build.createdAt build.finishedAt ] ]
    , row [] [ column [] [ text <| "Started build " ++ timeDistanceWithAffix model.time build.createdAt ] ]
    , row [ width shrink, centerX ] [ column [ padding 20 ] [ pauseButton model.buildLogPaused build.id ], column [] [ restartButton build.id ] ]
    , row [ width (fill |> maximum 1000), height (fill |> maximum 1000) ]
        [ column
            [ width (fill |> maximum 1000)
            , height (fill |> maximum 500)
            , scrollbars
            , spacing 4
            ]
          <|
            viewLogLines (formatTime build.createdAt) build.log
        ]
    ]


viewLogLines : (Time.Posix -> String) -> List LogLine -> List (Element Msg)
viewLogLines timeFormat lines =
    let
        listLength =
            List.length lines - 1
    in
    List.indexedMap (\idx line -> viewLogLine (idx == listLength) timeFormat line) lines


viewLogLine : Bool -> (Time.Posix -> String) -> LogLine -> Element Msg
viewLogLine lastLogLine timeFormat line =
    let
        additional =
            if lastLogLine then
                [ htmlAttribute (id "last-log-line") ]

            else
                []
    in
    row ([ width fill, spacing 8 ] ++ additional)
        [ column [ width shrink, Font.color (rgba 1 1 1 0.6) ] [ text <| timeFormat line.createdAt ]
        , column [ width shrink ] [ text line.line ]
        ]


formatTime : Time.Posix -> Time.Posix -> String
formatTime buildStart logTime =
    let
        diff =
            Time.millisToPosix <| Time.posixToMillis logTime - Time.posixToMillis buildStart

        seconds =
            Time.toSecond Time.utc diff

        minutes =
            Time.toMinute Time.utc diff

        hours =
            Time.toHour Time.utc diff
    in
    String.join ":" <|
        List.map (String.padLeft 2 '0' << String.fromInt)
            [ hours, minutes, seconds ]


pauseButton : Bool -> Int -> Element Msg
pauseButton paused id =
    if paused then
        styledButton { label = text "Continue", onPress = Just (BuildLogWatch id) }

    else
        styledButton { label = text "Pause", onPress = Just (BuildLogUnwatch id) }


restartButton : Int -> Element Msg
restartButton id =
    styledButton { label = text "Restart this build", onPress = Just (BuildRestart id) }
