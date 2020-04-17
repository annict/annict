# frozen_string_literal: true

module Db
  class PublishingStateLabelComponent < ApplicationComponent
    def initialize(resource:)
      @resource = resource
    end

    def call
      Htmlrb.build do |el|
        el.div class: label_class do
          label_text
        end
      end.html_safe
    end

    private

    attr_reader :resource

    def label_class
      resource.published? ? "badge badge-success" : "badge badge-warning"
    end

    def label_text
      I18n.t("resources.series.state.#{resource.published? ? 'published' : 'hidden'}")
    end
  end
end
