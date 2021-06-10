# frozen_string_literal: true

module ButtonGroups
  class EpisodeTransitionButtonGroupComponent < ApplicationV6Component
    def initialize(view_context, episode:)
      super view_context
      @episode = episode
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group btn-group-sm text-center" do
          h.tag :a, href: prev_episode_path, class: prev_episode_class_name do
            h.tag :i, class: "far fa-angle-left me-1"
            h.text t("noun.prev")
          end

          h.tag :a, href: next_episode_path, class: next_episode_class_name do
            h.text t("noun.next")
            h.tag :i, class: "far fa-angle-right ms-1"
          end
        end
      end
    end

    private

    def prev_episode_class_name
      class_name = %w[btn btn-secondary]

      if @episode.prev_episode.blank?
        class_name << "disabled"
      end

      class_name.join(" ")
    end

    def next_episode_class_name
      class_name = %w[btn btn-secondary]

      if @episode.next_episode.blank?
        class_name << "disabled"
      end

      class_name.join(" ")
    end

    def prev_episode_path
      if @episode.prev_episode.blank?
        return "#"
      end

      view_context.episode_path(@episode.anime.id, @episode.prev_episode.id)
    end

    def next_episode_path
      if @episode.next_episode.blank?
        return "#"
      end

      view_context.episode_path(@episode.anime.id, @episode.next_episode.id)
    end
  end
end
