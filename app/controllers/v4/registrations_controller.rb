# frozen_string_literal: true

module V4
  class RegistrationsController < V4::ApplicationController
    layout "simple"

    def new
      token = params[:token]

      unless token
        return redirect_to root_path
      end

      @session_interaction = SessionInteraction.find_by(kind: :sign_up, token: token)

      if !@session_interaction || @session_interaction.expired?
        @expired = true
        return
      end

      @session_interaction.touch(:expires_at)

      @form = RegistrationForm.new
      @form.email = @session_interaction.email
      @form.token = @session_interaction.token
    end

    def create
      token = registration_form_attributes[:token]
      @session_interaction = SessionInteraction.find_by(kind: :sign_up, token: token)

      unless @session_interaction
        return redirect_to root_path
      end

      @form = RegistrationForm.new(registration_form_params)

      return render(:new) unless @form.valid?

      user = User.new(
        username: @form.username,
        email: @form.email
      ).build_relations
      user.time_zone = cookies["ann_time_zone"].presence || "Asia/Tokyo"
      user.locale = locale
      user.confirmed_at = Time.zone.now
      user.setting.privacy_policy_agreed = true

      ActiveRecord::Base.transaction do
        user.save!
        @session_interaction.destroy
      end

      sign_in user

      flash[:notice] = t("messages.registrations.create.welcome")
      redirect_to root_path
    end

    private

    def registration_form_attributes
      @registration_form_attributes ||= params.to_unsafe_h["registration_form"].except(:email)
    end

    def registration_form_params
      RegistrationContract.new.call(registration_form_attributes.merge(email: @session_interaction.email))
    end
  end
end
