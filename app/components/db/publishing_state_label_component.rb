# typed: false
# frozen_string_literal: true

module Db
  class PublishingStateLabelComponent < Db::ApplicationComponent
    def initialize(resource:)
      @resource = resource
    end

    private

    attr_reader :resource

    def label_class
      resource.published? ? "badge bg-success" : "badge bg-warning"
    end

    def label_text
      I18n.t("resources.series.state.#{resource.published? ? "published" : "hidden"}")
    end
  end
end
