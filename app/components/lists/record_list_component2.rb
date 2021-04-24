# frozen_string_literal: true

module Lists
  class RecordListComponent2 < ApplicationComponent2
    def initialize(view_context, records:, show_card: true)
      super view_context
      @records = records
      @show_card = show_card
    end

    def render
      build_html do |h|
        if @records.present?
          @records.each do |record|
            h.tag :div, class: "py-3 u-underline" do
              h.tag :turbo_frame, id: dom_id(record) do
                h.tag :div, class: "mb-3" do
                  h.html RecordHeaderComponent2.new(view_context, record: record).render
                end

                if record.episode_record?
                  h.html EpisodeRecordContentComponent2.new(view_context, record: record, show_card: @show_card).render
                elsif record_entity.anime_record?
                  render AnimeRecordContentComponent.new(
                    user_entity: record_entity.user,
                    anime_entity: record_entity.trackable,
                    record_entity: record_entity,
                    anime_record_entity: record_entity.recordable,
                    show_card: @show_card
                  )
                end
              end
            end
          end

          if @records.respond_to?(:total_pages) && @records.total_pages > 1
            h.tag :div, class: "mt-3 text-center" do
              h.html paginate(@records)
            end
          end
        else
          h.html EmptyComponent2.new(view_context, text: t("messages._empty.no_records")).render
        end
      end
    end
  end
end
