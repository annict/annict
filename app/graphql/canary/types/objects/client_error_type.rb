# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ClientErrorType < Canary::Types::Objects::Base
        field :message, String, null: false
      end
    end
  end
end
