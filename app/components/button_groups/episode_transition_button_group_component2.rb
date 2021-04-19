# frozen_string_literal: true

module ButtonGroups
  class EpisodeTransitionButtonGroupComponent2 < ApplicationComponent
    def initialize(episode:)
      @episode = episode
    end

    private

    def prev_episode_class_name
      class_name = %w(btn btn-secondary)

      if @episode.prev_episode.blank?
        class_name << "disabled"
      end

      class_name.join(" ")
    end

    def next_episode_class_name
      class_name = %w(btn btn-secondary)

      if @episode.next_episode.blank?
        class_name << "disabled"
      end

      class_name.join(" ")
    end

    def prev_episode_path
      if @episode.prev_episode.blank?
        return "#"
      end

      episode_path(@episode.anime.id, @episode.prev_episode.id)
    end

    def next_episode_path
      if @episode.next_episode.blank?
        return "#"
      end

      episode_path(@episode.anime.id, @episode.next_episode.id)
    end
  end
end
