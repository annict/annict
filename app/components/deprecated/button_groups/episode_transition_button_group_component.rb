# frozen_string_literal: true

module Deprecated::ButtonGroups
  class EpisodeTransitionButtonGroupComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, episode:, path: :episode_path, button_style: true, class_name: "")
      super view_context
      @episode = episode
      @path = path
      @button_style = button_style
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: component_class_name do
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

    def component_class_name
      class_name = @class_name.split(" ")
      class_name += %w[btn-group btn-group-sm text-center] if @button_style
      class_name.join(" ")
    end

    def prev_episode_class_name
      class_name = []
      class_name += %w[btn btn-secondary] if @button_style

      if @episode.prev_episode.blank?
        class_name << (@button_style ? "disabled" : "pe-none text-muted")
      end

      class_name.join(" ")
    end

    def next_episode_class_name
      class_name = []
      class_name += if @button_style
        %w[btn btn-secondary]
      else
        %w[ps-2]
      end

      if @episode.next_episode.blank?
        class_name << (@button_style ? "disabled" : "pe-none text-muted")
      end

      class_name.join(" ")
    end

    def prev_episode_path
      if @episode.prev_episode.blank?
        return "#"
      end

      case @path
      when :episode_path
        view_context.episode_path(@episode.work.id, @episode.prev_episode.id)
      when :fragment_trackable_episode_path
        view_context.fragment_trackable_episode_path(@episode.prev_episode.id)
      end
    end

    def next_episode_path
      if @episode.next_episode.blank?
        return "#"
      end

      case @path
      when :episode_path
        view_context.episode_path(@episode.work.id, @episode.next_episode.id)
      when :fragment_trackable_episode_path
        view_context.fragment_trackable_episode_path(@episode.next_episode.id)
      end
    end
  end
end
