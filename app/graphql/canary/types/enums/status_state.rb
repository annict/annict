# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class StatusState < Canary::Types::Enums::Base
        value "WANNA_WATCH", "見たい"
        value "WATCHING", "見てる"
        value "WATCHED", "見た"
        value "ON_HOLD", "一時中断"
        value "STOP_WATCHING", "視聴中止"
        value "NO_STATE", "未設定"
      end
    end
  end
end
