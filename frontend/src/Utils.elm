port module Utils exposing (scrollIntoView, timeDistanceWithAffix, timeDistanceWithoutAffix)

import Time
import Time.Distance
import Time.Distance.I18n


timeDistanceWithoutAffix : Time.Posix -> Time.Posix -> String
timeDistanceWithoutAffix fromTime toTime =
    Time.Distance.inWordsWithConfig { withAffix = False } Time.Distance.I18n.en toTime fromTime


timeDistanceWithAffix : Time.Posix -> Time.Posix -> String
timeDistanceWithAffix fromTime toTime =
    Time.Distance.inWordsWithConfig { withAffix = True } Time.Distance.I18n.en toTime fromTime


port scrollIntoView : String -> Cmd msg
