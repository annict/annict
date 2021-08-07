# frozen_string_literal: true

module Canary
  module Types
    module Interfaces
      module Reactable
        include Canary::Types::Interfaces::Base

        field :reactions, Canary::Types::Objects::ReactionType.connection_type, null: false
      end
    end
  end
end
