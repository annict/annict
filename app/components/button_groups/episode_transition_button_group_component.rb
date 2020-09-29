# frozen_string_literal: true

module ButtonGroups
  class EpisodeTransitionButtonGroupComponent < ApplicationComponent
    def initialize(episode_entity:)
      @episode_entity = episode_entity
    end

    private

    attr_reader :episode_entity

    def prev_episode_class_name
      class_name = %w(btn btn-secondary)

      if episode_entity.prev_episode.blank?
        class_name << "disabled"
      end

      class_name.join(" ")
    end

    def next_episode_class_name
      class_name = %w(btn btn-secondary)

      if episode_entity.next_episode.blank?
        class_name << "disabled"
      end

      class_name.join(" ")
    end

    def prev_episode_path
      if episode_entity.prev_episode.blank?
        return "#"
      end

      episode_path(episode_entity.anime.database_id, episode_entity.prev_episode.database_id)
    end

    def next_episode_path
      if episode_entity.next_episode.blank?
        return "#"
      end

      episode_path(episode_entity.anime.database_id, episode_entity.next_episode.database_id)
    end
  end
end
