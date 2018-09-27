#!/usr/bin/env ruby

require 'json'

ejson = JSON.parse($<.read)
# FIXME: kubernetes_secrets shouldn't be hardcoded
data = ejson['kubernetes_secrets']['credentials']['data']
puts data.transform_values { |value| [value].pack('m0') }.to_json
