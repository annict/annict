# frozen_string_literal: true

module ListGroups
  class RecordListGroupComponent < ApplicationV6Component
    def initialize(view_context, my_records:, following_records:, all_records:)
      super view_context
      @my_records = my_records
      @following_records = following_records
      @all_records = all_records
    end

    def render
      build_html do |h|
        h.tag :div, class: "container" do
          h.tag :h3, class: "fw-bold mb-0" do
            h.text t("noun.my_records")
          end
        end

        h.tag :div, class: "container mt-3 u-container-flat" do
          h.html Lists::RecordListComponent.new(
            view_context,
            records: @my_records,
            show_box: false
          ).render
        end

        h.tag :div, class: "container mt-5" do
          h.tag :h3, class: "fw-bold mb-0" do
            h.text t("noun.following_records")
          end
        end

        h.tag :div, class: "container mt-3 u-container-flat" do
          h.html Lists::RecordListComponent.new(
            view_context,
            records: @following_records,
            show_box: false
          ).render
        end

        h.tag :div, class: "container mt-5" do
          h.tag :h3, class: "fw-bold mb-0" do
            h.text t("noun.other_comments")
          end
        end

        h.tag :div, class: "container mt-3 u-container-flat" do
          h.html Lists::RecordListComponent.new(
            view_context,
            records: @all_records,
            show_box: false,
            empty_text: :no_comments
          ).render
        end
      end
    end
  end
end
