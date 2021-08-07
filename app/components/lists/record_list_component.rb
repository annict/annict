# frozen_string_literal: true

module Lists
  class RecordListComponent < ApplicationV6Component
    def initialize(view_context, records:, show_box: true, empty_text: :no_records)
      super view_context
      @records = records
      @show_box = show_box
      @empty_text = empty_text
      @pagenation = @records.respond_to?(:first_page?)
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-list" do
          if @records.present?
            @records.each do |record|
              h.tag :div, class: "card mt-3 u-card-flat" do
                h.tag :div, class: "card-body" do
                  h.html ListItems::RecordListItemComponent.new(view_context, record: record, show_box: @show_box).render
                end
              end
            end

            if @pagenation
              h.tag :div, class: "mt-3 text-center" do
                h.html ButtonGroups::PaginationButtonGroupComponent.new(view_context, collection: @records).render
              end
            end
          else
            h.tag :div, class: "card mt-3 u-card-flat" do
              h.tag :div, class: "card-body" do
                h.html EmptyV6Component.new(view_context, text: t("messages._empty.#{@empty_text}")).render
              end
            end
          end
        end
      end
    end
  end
end
