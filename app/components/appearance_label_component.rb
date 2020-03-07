# frozen_string_literal: true

class AppearanceLabelComponent < ApplicationComponent
  inline!

  def initialize(resource:)
    @resource = resource
  end

  def call
    content_tag :span, class: "badge" do
      label_text
    end
  end

  private

  attr_reader :resource

  def label_text
    resource.disappeared_at ? t("noun.disappeared") : t("noun.appeared")
  end
end
