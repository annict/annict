# frozen_string_literal: true

module Types
  module InputObjects
    class EpisodeOrder < Types::InputObjects::Base
      argument :field, Types::Enums::EpisodeOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
