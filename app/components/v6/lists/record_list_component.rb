# frozen_string_literal: true

module V6::Lists
  class RecordListComponent < V6::ApplicationComponent
    def initialize(view_context, records:, show_box: true)
      super view_context
      @records = records
      @show_box = show_box
    end

    def render
      build_html do |h|
        if @records.present?
          @records.each do |record|
            h.tag :div, class: "py-3 u-underline" do
              h.tag :turbo_frame, id: dom_id(record) do
                h.tag :div, class: "mb-3" do
                  h.html V6::RecordHeaderComponent.new(view_context, record: record).render
                end

                if record.episode_record?
                  h.html V6::Contents::EpisodeRecordContentComponent.new(
                    view_context,
                    record: record,
                    show_box: @show_box
                  ).render
                elsif record.anime_record?
                  h.html V6::Contents::AnimeRecordContentComponent.new(
                    view_context,
                    record: record,
                    show_box: @show_box
                  ).render
                end
              end
            end
          end

          h.tag :div, class: "mt-3 text-center" do
            h.html V6::ButtonGroups::PaginationButtonGroupComponent.new(view_context, collection: @records).render
          end
        else
          h.html V6::EmptyComponent.new(view_context, text: t("messages._empty.no_records")).render
        end
      end
    end
  end
end
