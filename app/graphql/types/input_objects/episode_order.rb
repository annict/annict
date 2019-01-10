# frozen_string_literal: true

module Types
  module InputObjects
    class EpisodeOrder < Types::InputObjects::Base
      argument :field, Types::Enum::EpisodeOrderField, required: true
      argument :direction, !Types::Enum::OrderDirection, required: true
    end
  end
end
