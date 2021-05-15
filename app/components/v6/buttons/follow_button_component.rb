# frozen_string_literal: true

module V6::Buttons
  class FollowButtonComponent < V6::ApplicationComponent
    def initialize(view_context, user:, page_category:, class_name: "")
      super view_context
      @user = user
      @page_category = page_category
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: "btn btn-primary" do
          h.text "Follow"
        end
      end
    end
  end
end
