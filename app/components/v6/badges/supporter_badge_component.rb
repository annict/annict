# frozen_string_literal: true

module V6::Badges
  class SupporterBadgeComponent < V6::ApplicationComponent
    def initialize(view_context, user:)
      super view_context
      @user = user
    end

    def render
      return "" unless @user.supporter?

      build_html do |h|
        h.tag :div, class: "badge u-bg-supporter" do
          h.text t("noun.supporter")
        end
      end
    end
  end
end
