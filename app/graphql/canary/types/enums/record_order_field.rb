# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class RecordOrderField < Canary::Types::Enums::Base
        value "CREATED_AT", "作成日時"
        value "LIKES_COUNT", "いいね数"
        value "RATING", "評価"
      end
    end
  end
end
