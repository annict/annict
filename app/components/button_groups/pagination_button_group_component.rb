# frozen_string_literal: true

module ButtonGroups
  class PaginationButtonGroupComponent < ApplicationV6Component
    def initialize(view_context, collection:, controller: nil, action: nil)
      super view_context
      @collection = collection
      @controller = controller
      @action = action
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group" do
          h.tag :a, class: prev_button_class_name, href: prev_page_path do
            h.tag :i, class: "far fa-angle-left me-2"
            h.text t("noun.prev_page")
          end

          h.tag :a, class: next_button_class_name, href: next_page_path do
            h.text t("noun.next_page")
            h.tag :i, class: "far fa-angle-right ms-2"
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
      %w[btn-outline-secondary disabled]
    end

    def prev_button_class_name
      class_name = base_button_class_name

      class_name += if @collection.first_page?
        disabled_button_class_name
      else
        enabled_button_class_name
      end

      class_name.join(" ")
    end

    def next_button_class_name
      class_name = base_button_class_name

      class_name += if @collection.last_page?
        disabled_button_class_name
      else
        enabled_button_class_name
      end

      class_name.join(" ")
    end

    def controller
      @controller.presence || view_context.params[:controller]
    end

    def action
      @action.presence || view_context.params[:action]
    end

    def prev_page_path
      @prev_page_path ||= view_context.prev_page_path(@collection, params: {controller: controller, action: action})
    end

    def next_page_path
      @next_page_path ||= view_context.next_page_path(@collection, params: {controller: controller, action: action})
    end
  end
end
