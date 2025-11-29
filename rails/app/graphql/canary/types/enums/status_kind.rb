# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class StatusKind < Canary::Types::Enums::Base
        value "PLAN_TO_WATCH", "見たい"
        value "WATCHING", "見てる"
        value "COMPLETED", "見た"
        value "ON_HOLD", "一時中断"
        value "DROPPED", "視聴中止"
        value "NO_STATUS", "未設定"
      end
    end
  end
end
