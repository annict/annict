class UsersController < ApplicationController
  permits :email

  before_filter :authenticate_user!, only: [:update]
  before_filter :set_user, only: [:show, :works]


  def show
    @works = @user.watching_works.order(released_at: :desc)
  end

  def works(status_kind, page)
    @works = @user.works_on(status_kind).page(page)
  end

  def update(user)
    current_user.email = user[:email]

    if current_user.valid?
      current_user.update_column(:unconfirmed_email, user[:email])
      current_user.resend_confirmation_instructions
      redirect_to setting_path, notice: t('registrations.create.confirmation_mail_has_sent')
    else
      render '/settings/show'
    end
  end


  private

  def set_user
    @user = User.find_by!(username: params[:id])
  end
end
