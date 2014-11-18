class UsersController < ApplicationController
  permits :email

  before_filter :authenticate_user!, only: [:update, :share]
  before_filter :set_user, only: [:show, :works]


  def show
    @watching_works = @user.watching_works
    checkined_works = @watching_works.checkined(@user).order('checkins.id DESC')
    other_works = @watching_works.where.not(id: checkined_works.pluck(:id))
    @works = (checkined_works + other_works).first(9)
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

  def share(body)
    TwitterWatchingShareWorker.perform_async(current_user.id, body)

    redirect_to user_path(current_user.username), notice: 'ツイートしました。'
  end


  private

  def set_user
    @user = User.find_by!(username: params[:id])
  end
end
