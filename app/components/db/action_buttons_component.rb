# frozen_string_literal: true

module Db
  class ActionButtonsComponent < ApplicationComponent
    def initialize(user:, resource:, detail_path:, edit_path:, publishing_path:)
      @user = user
      @resource = resource
      @detail_path = detail_path
      @edit_path = edit_path
      @publishing_path = publishing_path
    end

    def call
      Htmlrb.build do |el|
        if series_policy.edit?
          el.div class: "mb-1" do
            el.a class: "btn btn-primary btn-sm", href: edit_path do
              I18n.t("noun.edit")
            end
          end
        end

        if series_policy.publish?
          el.div class: "mb-1" do
            el.a class: "btn btn-warning btn-sm", data_method: publishing_method, data_confirm: t("messages._common.are_you_sure"), href: publishing_path do
              publishing_text
            end
          end
        end

        if series_policy.destroy?
          el.div class: "mb-1" do
            el.a class: "btn btn-danger btn-sm", data_method: "delete", data_confirm: t("messages._common.are_you_sure"), href: detail_path do
              I18n.t("noun.delete")
            end
          end
        end
      end.html_safe
    end

    private

    attr_reader :detail_path, :edit_path, :publishing_path, :resource, :user

    def series_policy
      SeriesPolicy.new(user, resource)
    end

    def publishing_method
      resource.published? ? "delete" : "post"
    end

    def publishing_text
      resource.published? ? I18n.t("noun.unpublish") : I18n.t("noun.publish")
    end
  end
end
