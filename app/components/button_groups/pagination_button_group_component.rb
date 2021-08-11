# frozen_string_literal: true

module ButtonGroups
  class PaginationButtonGroupComponent < ApplicationV6Component
    PAGE_WINDOW = 10
    PARAM_KEY_EXCEPT_LIST = Kaminari::Helpers::PARAM_KEY_EXCEPT_LIST

    def initialize(view_context, collection:, without_count: false, controller: nil, action: nil)
      super view_context
      @collection = collection
      @without_count = without_count
      @controller = controller
      @action = action
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group c-pagination-button-group" do
          unless @without_count
            h.tag :a, class: prev_button_class_name, href: page_url_for(page: 1) do
              h.tag :i, class: "far fa-angle-double-left"
            end
          end

          h.tag :a, class: prev_button_class_name, href: page_url_for(page: current_page - 1) do
            h.tag :i, class: "far fa-angle-left"
          end

          unless @without_count
            h.tag :div, class: "btn-group" do
              h.tag :button, class: "#{prev_button_class_name} dropdown-toggle", id: "prevDrop", type: "button", data_bs_toggle: "dropdown" do
                h.text "..."
              end

              h.tag :ul, class: "dropdown-menu" do
                prev_pages.each do |page|
                  h.tag :li do
                    h.tag :a, class: "dropdown-item", href: page_url_for(page: page) do
                      h.text page
                    end
                  end
                end
              end
            end
          end

          h.tag :a, class: "active btn btn-secondary disabled", href: "#" do
            h.text current_page
          end

          unless @without_count
            h.tag :div, class: "btn-group" do
              h.tag :button, class: "#{next_button_class_name} dropdown-toggle", id: "nextDrop", type: "button", data_bs_toggle: "dropdown" do
                h.text "..."
              end

              h.tag :ul, class: "dropdown-menu" do
                next_pages.each do |page|
                  h.tag :li do
                    h.tag :a, class: "dropdown-item", href: page_url_for(page: page) do
                      h.text page
                    end
                  end
                end
              end
            end
          end

          h.tag :a, class: next_button_class_name, href: page_url_for(page: current_page + 1) do
            h.tag :i, class: "far fa-angle-right"
          end

          unless @without_count
            h.tag :a, class: next_button_class_name, href: page_url_for(page: total_page) do
              h.tag :i, class: "far fa-angle-double-right"
            end
          end
        end
      end
    end

    private

    def base_button_class_name
      %w[btn]
    end

    def enabled_button_class_name
      %w[btn-secondary]
    end

    def disabled_button_class_name
      %w[btn-outline-secondary disabled opacity-25]
    end

    def prev_button_class_name
      @prev_button_class_name ||= begin
        class_name = base_button_class_name

        class_name += if @collection.first_page?
          disabled_button_class_name
        else
          enabled_button_class_name
        end

        class_name.join(" ")
      end
    end

    def next_button_class_name
      @next_button_class_name ||= begin
        class_name = base_button_class_name

        class_name += if @collection.last_page?
          disabled_button_class_name
        else
          enabled_button_class_name
        end

        class_name.join(" ")
      end
    end

    def controller
      @controller.presence || view_context.params[:controller]
    end

    def action
      @action.presence || view_context.params[:action]
    end

    def page_url_for(page:)
      params = view_context.params.to_unsafe_h.except(*PARAM_KEY_EXCEPT_LIST)
      params[:page] = page
      params[:controller] = controller
      params[:action] = action
      params[:only_path] = true

      url_for params
    end

    def current_page
      @collection.current_page
    end

    def total_page
      @collection.total_pages
    end

    def next_pages
      @next_pages ||= (1..PAGE_WINDOW / 2).to_a.map { |n| current_page + n }.select { |n| n <= total_page }
    end

    def prev_pages
      @prev_pages ||= (1..PAGE_WINDOW / 2).to_a.reverse.map { |n| current_page - n }.select { |n| n.positive? }
    end
  end
end
