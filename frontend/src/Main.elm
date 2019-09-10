module Main exposing (main)

import Browser
import Init exposing (init)
import Model exposing (Model)
import Msgs exposing (Msg(..))
import Platform.Sub exposing (batch)
import Time
import Update exposing (update)
import View exposing (view)
import WebSocket as WS


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        , onUrlRequest = LinkClicked
        , onUrlChange = UrlChanged
        }


subscriptions : Model -> Sub Msg
subscriptions _ =
    batch
        [ WS.subscriptions
        , Time.every 60000 Tick
        ]
