# frozen_string_literal: true

module Db
  class PeopleController < Db::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update hide destroy)

    def index
      @people = Person.order(id: :desc).page(params[:page])
    end

    def new
      @person = Person.new
      authorize @person, :new?
    end

    def create
      @person = Person.new(person_params)
      @person.user = current_user
      authorize @person, :create?

      return render(:new) unless @person.valid?
      @person.save_and_create_activity!

      redirect_to db_people_path, notice: t("resources.person.created")
    end

    def edit
      @person = Person.find(params[:id])
      authorize @person, :edit?
    end

    def update
      @person = Person.find(params[:id])
      authorize @person, :update?

      @person.attributes = person_params
      @person.user = current_user

      return render(:edit) unless @person.valid?
      @person.save_and_create_activity!

      redirect_to edit_db_person_path(@person), notice: t("resources.person.updated")
    end

    def hide
      @person = Person.find(params[:id])
      authorize @person, :hide?

      @person.soft_delete

      flash[:notice] = t("resources.person.unpublished")
      redirect_back fallback_location: db_people_path
    end

    def destroy
      @person = Person.find(params[:id])
      @person.destroy

      flash[:notice] = t("resources.person.deleted")
      redirect_back fallback_location: db_people_path
    end

    def activities
      @person = Person.find(params[:id])
      @activities = @person.db_activities.order(id: :desc)
      @comment = @person.db_comments.new
    end

    private

    def person_params
      params.require(:person).permit(
        :name, :name_kana, :name_en, :nickname, :nickname_en, :gender,
        :blood_type, :prefecture_id, :birthday, :height, :url, :url_en,
        :wikipedia_url, :wikipedia_url_en, :twitter_username, :twitter_username_en
      )
    end
  end
end
