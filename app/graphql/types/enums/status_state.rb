# frozen_string_literal: true

module Types
  module Enums
    class StatusState < Types::Enums::Base
      value "WANNA_WATCH", ""
      value "WATCHING", ""
      value "WATCHED", ""
      value "ON_HOLD", ""
      value "STOP_WATCHING", ""
      value "NO_STATE", ""
    end
  end
end
