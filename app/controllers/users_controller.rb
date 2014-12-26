class UsersController < ApplicationController
  permits :email

  before_filter :authenticate_user!, only: [:update]
  before_filter :set_user, only: [:show, :works]


  def show
    @watching_works = @user.works.watching
    checkedin_works = @watching_works.checkedin_by(@user).order('c2.checkin_id DESC')
    other_works = @watching_works.where.not(id: checkedin_works.pluck(:id))
    @works = (checkedin_works + other_works).first(9)
  end

  def works(status_kind, page: nil)
    @works = @user.works.on(status_kind).order(released_at: :desc).page(page)
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
