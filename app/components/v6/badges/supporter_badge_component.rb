# frozen_string_literal: true

module Badges
  class SupporterBadgeComponent < ApplicationComponent
    def initialize(view_context, user:)
      super view_context
      @user = user
    end

    def render
      return "" unless @user.supporter?

      build_html do |h|
        h.tag :div, class: "badge bg-supporter" do
          h.text t("noun.supporter")
        end
      end
    end
  end
end
