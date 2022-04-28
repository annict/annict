# frozen_string_literal: true

module Boxes
  class WorkBoxComponent < ApplicationV6Component
    def initialize(view_context, work:, episode: nil)
      super view_context
      @work = work
      @episode = episode
    end

    def render
      build_html do |h|
        h.tag :div, class: "row g-3" do
          h.tag :div, class: "col-auto" do
            h.tag :a, href: view_context.work_path(@work.id), target: "_top" do
              h.html view_context.render(
                Pictures::WorkPictureComponent.new(
                  work: @work,
                  width: 80
                )
              )
            end
          end

          h.tag :div, class: "col" do
            h.tag :div do
              h.tag :a, href: view_context.work_path(@work.id), class: "text-body", target: "_top" do
                h.text @work.local_title
              end
            end

            if @episode
              h.tag :div do
                h.tag :a, href: view_context.episode_path(@work.id, @episode.id), class: "fw-bold small text-body", target: "_top" do
                  h.tag :span, class: "px-1" do
                    h.text @episode.local_number
                  end

                  h.text @episode.local_title
                end
              end
            end

            h.tag :div, class: "mt-1" do
              h.html ButtonGroups::WorkButtonGroupComponent.new(view_context, work: @work).render
            end
          end
        end
      end
    end
  end
end
