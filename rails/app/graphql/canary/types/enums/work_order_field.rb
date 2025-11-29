# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class WorkOrderField < Canary::Types::Enums::Base
        value "CREATED_AT", "作成日時"
        value "SEASON", "シーズン"
        value "WATCHERS_COUNT", "視聴者数"
      end
    end
  end
end
