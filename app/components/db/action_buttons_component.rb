# frozen_string_literal: true

module Db
  class ActionButtonsComponent < Db::ApplicationComponent
    def initialize(user:, resource:, detail_path:, edit_path:, publishing_path:)
      @user = user
      @resource = resource
      @detail_path = detail_path
      @edit_path = edit_path
      @publishing_path = publishing_path
    end

    private

    attr_reader :detail_path, :edit_path, :publishing_path, :resource, :user

    def db_resource_policy
      @db_resource_policy ||= Pundit::PolicyFinder.new(resource).policy.new(user, resource)
    end

    def publishing_method
      resource.published? ? "delete" : "post"
    end

    def publishing_btn_class
      resource.published? ? "btn-warning" : "btn-success"
    end

    def publishing_text
      resource.published? ? I18n.t("noun.unpublish") : I18n.t("noun.publish")
    end
  end
end
