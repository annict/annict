# frozen_string_literal: true

module V6::Headers
  class ProfileHeaderComponent < V6::ApplicationComponent
    def initialize(view_context, user:, current_user:, params:)
      super view_context
      @user = user
      @current_user = current_user
      @params = params
      @profile = @user.profile
    end

    def render
      build_html do |h|
        h.tag :div, class: "align-items-center row gx-3" do
          h.tag :div, class: "col-auto" do
            h.tag :a, href: view_context.profile_path(@user.username) do
              h.html V6::Pictures::AvatarPictureComponent.new(view_context, user: @user, width: 80, mb_width: 80).render
            end
          end

          h.tag :div, class: "col" do
            if @user.supporter?
              h.html V6::Badges::SupporterBadgeComponent.new(view_context, user: @user).render
            end

            h.tag :h1, class: "h2" do
              h.tag :a, class: "text-body", href: view_context.profile_path(@user.username) do
                h.text @profile.name

                h.tag :div, class: "h4 text-muted" do
                  h.text "@#{@user.username}"
                end
              end
            end
          end

          h.tag :div, class: "col-12 col-sm-auto text-center" do
            if current_user&.id == @user.id
              h.tag :a, href: view_context.settings_profile_path, class: "btn btn-outline-secondary" do
                h.text t("noun.edit_profile")
              end
            else
              h.html V6::Buttons::FollowButtonComponent.new(view_context, user: @user, page_category: page_category).render
            end
          end
        end

        h.tag :div, class: "c-nav" do
          h.tag :ul, class: "c-nav__list" do
            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.profile"), view_context.profile_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "v6/users" && @params[:action] == "show"
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.records"), view_context.record_list_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "v4/records" && @params[:action].in?(%w[index show])
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
                active: @params[:controller].in?(%w[v3/favorite_characters favorite_people favorite_organizations]) && @params[:action] == "index"
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.following"), view_context.following_user_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "users" && @params[:action] == "following"
            end

            h.tag :li, class: "c-nav__item" do
              h.html active_link_to t("noun.followers"), view_context.followers_user_path(@user.username),
                class: "c-nav__link",
                class_active: "c-nav__link--active",
                active: @params[:controller] == "users" && @params[:action] == "followers"
            end
          end
        end
      end
    end
  end
end
