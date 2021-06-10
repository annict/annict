# frozen_string_literal: true

module ButtonGroups
  class PaginationButtonGroupComponent < ApplicationV6Component
    def initialize(view_context, collection:)
      super view_context
      @collection = collection
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn-group" do
          h.tag :a, class: prev_button_class_name, href: view_context.path_to_prev_page(@collection) do
            h.tag :i, class: "far fa-angle-left me-2"
            h.text t("noun.prev_page")
          end

          h.tag :a, class: next_button_class_name, href: view_context.path_to_next_page(@collection) do
            h.text t("noun.next_page")
            h.tag :i, class: "far fa-angle-right ms-2"
          end
        end
      end
    end

    private

    def base_button_class_name
      %w[btn btn-secondary]
    end

    def prev_button_class_name
      class_name = base_button_class_name
      class_name << "disabled" if @collection.first_page?
      class_name.join(" ")
    end

    def next_button_class_name
      class_name = base_button_class_name
      class_name << "disabled" if @collection.last_page?
      class_name.join(" ")
    end
  end
end
