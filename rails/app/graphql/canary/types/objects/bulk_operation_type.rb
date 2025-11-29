# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class BulkOperationType < Canary::Types::Objects::Base
        field :job_id, String, null: false
      end
    end
  end
end
