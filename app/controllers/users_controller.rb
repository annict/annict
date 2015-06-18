class UsersController < ApplicationController
  permits :email

  before_filter :authenticate_user!, only: [:update, :destroy, :share]
  before_filter :set_user, only: [:show, :works, :following, :followers]


  def show
    @watching_works = @user.works.watching
    checkedin_works = @watching_works.checkedin_by(@user).order('c2.checkin_id DESC')
    other_works = @watching_works.where.not(id: checkedin_works.pluck(:id))
    @works = (checkedin_works + other_works).first(9)
    @graph_labels = Annict::Graphs::Checkins.labels
    @graph_values = Annict::Graphs::Checkins.values(@user)
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

  def share(body)
    TwitterWatchingShareWorker.perform_async(current_user.id, body)

    redirect_to user_path(current_user.username), notice: 'ツイートしました。'
  end

  def following
    @users = @user.followings.order('follows.id DESC')
  end

  def followers
    @users = @user.followers.order('follows.id DESC')
  end

  def destroy
    current_user.destroy
    redirect_to root_path, notice: "退会しました。(´・ω;:.."
  end

  private

  def set_user
    @user = User.find_by!(username: params[:id])
  end
end
