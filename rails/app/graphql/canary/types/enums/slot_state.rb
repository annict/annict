# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class SlotState < Canary::Types::Enums::Base
        value "PUBLISHED", "公開中"
        value "HIDDEN", "非公開中"
      end
    end
  end
end
