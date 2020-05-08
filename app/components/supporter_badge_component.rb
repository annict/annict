# frozen_string_literal: true

class SupporterBadgeComponent < ApplicationComponent
  def initialize(user_entity:)
    @user_entity = user_entity
  end

  def call
    return unless user_entity.display_supporter_badge

    content_tag :div, class: "badge u-badge-supporter" do
      I18n.t("noun.supporter")
    end
  end

  private

  attr_reader :user_entity
end
