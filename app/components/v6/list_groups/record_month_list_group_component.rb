# frozen_string_literal: true

module V6::ListGroups
  class RecordMonthListGroupComponent < V6::ApplicationComponent
    def initialize(view_context, user:, months:, controller_name:, current_month:)
      super view_context
      @user = user
      @months = months
      @controller_name = controller_name
      @current_month = current_month
    end

    def render
      build_html do |h|
        h.tag :div, class: "list-group" do
          h.tag :a, href: view_context.record_list_path(@user.username), class: all_link_class_name do
            h.text t("noun.all")
            h.tag :span, class: "badge badge-pill badge-primary" do
              h.text @months.values.reduce(&:+)
            end
          end

          @months.each do |month, count|
            if count > 0
              h.tag :a, href: view_context.record_list_path(@user.username, month: month.to_s(:ym)), class: month_link_class_name(month) do
                h.text month.to_s(:ym)

                h.tag :span, class: "badge badge-pill badge-primary" do
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

    def month_link_class_name(month)
      class_name = "align-items-center d-flex justify-content-between list-group-item"
      class_name += " active" if @current_month == month.to_s(:ym)
      class_name
    end
  end
end
