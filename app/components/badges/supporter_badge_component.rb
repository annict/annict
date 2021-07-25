# frozen_string_literal: true

module Badges
  class SupporterBadgeComponent < ApplicationV6Component
    def initialize(view_context, user:)
      super view_context
      @user = user
    end

    def render
      return "" unless @user.supporter?
      return "" if @user.supporter? && @user.setting.hide_supporter_badge?

      build_html do |h|
        h.tag :div, class: "badge u-bg-supporter" do
          h.text t("noun.supporter")
        end
      end
    end
  end
end
