module Data exposing (Build, LogLine, Organization, Socket)

import Time


type alias Socket =
    { url : String
    , fd : Int
    }


type alias Organization =
    { owner : String
    , url : String
    , buildCount : Int
    }


type alias LogLine =
    { buildId : Int
    , createdAt : Time.Posix
    , line : String
    }


type alias Build =
    { id : Int
    , status : String
    , createdAt : Time.Posix
    , finishedAt : Time.Posix
    , owner : String
    , ownerLink : String
    , repo : String
    , repoLink : String
    , sha : String
    , prNumber : Int
    , prTitle : String
    , branch : String
    , userLogin : String
    , log : List LogLine
    }
