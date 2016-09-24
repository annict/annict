# frozen_string_literal: true

module DB
  class EpisodeForm
    include ActiveModel::Model
    include Virtus.model

    attribute :raw_number, String
    attribute :number, String
    attribute :title, String
    attribute :title_ro, String
    attribute :title_en, String
    attribute :prev_episode_id, Integer
    attribute :sort_number, Integer
    attribute :sc_count, Integer
    attribute :fetch_syobocal, Boolean
  end
end
