# frozen_string_literal: true

module ListGroups
  class EpisodeRecordListGroupComponent < ApplicationComponent2
    def initialize(view_context, my_records:, following_records:, all_records:)
      super view_context
      @my_records = my_records
      @following_records = following_records
      @all_records = all_records
    end

    def render
      build_html do |h|
        h.tag :div, class: "mb-3" do
          h.tag :h3, class: "font-weight-bold mb-3" do
            h.text t("noun.my_records")
          end

          if current_user
            h.html Lists::RecordListComponent2.new(view_context, records: @my_records, show_card: false).render
          end
        end

        h.tag :hr, class: "mb-5"

        h.tag :div, class: "mb-3" do
          h.tag :h3, class: "font-weight-bold mb-3" do
            h.text t("noun.following_records")
          end

          if current_user
            h.html Lists::RecordListComponent2.new(view_context, records: @following_records, show_card: false).render
          end
        end

        h.tag :hr, class: "mb-5"

        h.tag :div, class: "mb-3" do
          h.tag :h3, class: "font-weight-bold mb-3" do
            h.text t("noun.other_comments")
          end

          h.html Lists::RecordListComponent2.new(view_context, records: @all_records, show_card: false).render
        end
      end
    end
  end
end
