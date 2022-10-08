# frozen_string_literal: true

module Deprecated::ListGroups
  class RecordMonthListGroupComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, user:, dates:, controller_name:, current_year:, current_month:)
      super view_context
      @user = user
      @dates = dates
      @controller_name = controller_name
      @current_year = current_year&.to_i
      @current_month = current_month&.to_i
    end

    def render
      build_html do |h|
        h.tag :div, class: "list-group rounded-3" do
          h.tag :a, href: view_context.record_list_path(@user.username), class: all_link_class_name do
            h.text t("noun.all")
            h.tag :span, class: "badge badge-pill bg-secondary" do
              h.text @dates.values.reduce(&:+)
            end
          end

          @dates.each do |date, count|
            if count > 0
              h.tag :a, href: view_context.record_list_path(@user.username, year: date.year, month: date.month), class: month_link_class_name(date) do
                h.text date.to_s(:ym)

                h.tag :span, class: "badge badge-pill bg-secondary" do
                  h.text count
                end
              end
            end
          end
        end
      end
    end

    private

    def all_link_class_name
      class_name = "align-items-center d-flex justify-content-between list-group-item"
      class_name += " active" if @controller_name == "v6/records" && @current_month.nil?
      class_name
    end

    def month_link_class_name(date)
      class_name = "align-items-center d-flex justify-content-between list-group-item"
      class_name += " active" if @current_year == date.year && @current_month == date.month
      class_name
    end
  end
end
