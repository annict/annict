# frozen_string_literal: true

module Headers
  class ProfileHeaderComponent < ApplicationV6Component
    def initialize(view_context, user:)
      super view_context
      @user = user
      @profile = @user.profile
    end

    def render
      build_html do |h|
        h.tag :div, class: "align-items-center row gx-3" do
          h.tag :div, class: "col-auto" do
            h.tag :a, href: view_context.profile_path(@user.username) do
              h.html view_context.render(Pictures::AvatarPictureComponent.new(user: @user, width: 80))
            end
          end

          h.tag :div, class: "col" do
            if @user.supporter? && !@user.setting.hide_supporter_badge?
              h.html Badges::SupporterBadgeComponent.new(view_context, user: @user).render
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
              h.html ButtonGroups::UserButtonGroupComponent.new(view_context, user: @user).render
            end
          end
        end
      end
    end
  end
end
