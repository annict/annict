# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module Enums
      class Media < Beta::Types::Enums::Base
        description "Media of anime"

        value "TV", ""
        value "OVA", ""
        value "MOVIE", ""
        value "WEB", ""
        value "OTHER", ""
      end
    end
  end
end
