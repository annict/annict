# frozen_string_literal: true

module Navs
  class UserNavComponent < ApplicationV6Component
    def initialize(view_context, user:, params:, class_name: "")
      super view_context
      @user = user
      @params = params
      @class_name = class_name
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-nav #{@class_name}" do
          h.tag :ul, class: "c-nav__list" do
            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.profile"), view_context.profile_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: [profiles: :show]
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.records"), view_context.record_list_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "records"
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.library"), view_context.library_path(@user.username, :watching),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "libraries" && @params[:action] == "show"
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.favorites"), view_context.favorite_character_list_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller].in?(%w[favorite_characters favorite_people favorite_organizations])
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.collections"), view_context.user_collection_list_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller].in?(%w[collection_items collections])
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.followees"), view_context.followee_list_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "followees"
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.followers"), view_context.follower_list_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "followers"
            end
          end
        end
      end
    end
  end
end
