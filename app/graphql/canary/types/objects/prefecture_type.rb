# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class PrefectureType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :name, String, null: false
      end
    end
  end
end
