# frozen_string_literal: true

module Beta
  module Types
    module Enums
      class StatusState < Beta::Types::Enums::Base
        value "WANNA_WATCH", ""
        value "WATCHING", ""
        value "WATCHED", ""
        value "ON_HOLD", ""
        value "STOP_WATCHING", ""
        value "NO_STATE", ""
      end
    end
  end
end
