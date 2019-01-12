# frozen_string_literal: true

module Types
  module Enums
    class Media < Types::Enums::Base
      description "Media of anime"

      value "TV", ""
      value "OVA", ""
      value "MOVIE", ""
      value "WEB", ""
      value "OTHER", ""
    end
  end
end
