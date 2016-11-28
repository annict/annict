# frozen_string_literal: true

module Db
  class PeopleController < Db::ApplicationController
    permits :name, :name_kana, :name_en, :nickname, :nickname_en, :gender,
      :blood_type, :prefecture_id, :birthday, :height, :url, :url_en,
      :wikipedia_url, :wikipedia_url_en, :twitter_username, :twitter_username_en

    before_action :authenticate_user!
    before_action :load_person, only: %i(edit update hide destroy activities)

    def index(page: nil)
      @people = Person.order(id: :desc).page(page)
    end

    def new
      @person = Person.new
      authorize @person, :new?
    end

    def create(person)
      @person = Person.new(person)
      @person.user = current_user
      authorize @person, :create?

      return render(:new) unless @person.valid?
      @person.save_and_create_activity!

      redirect_to edit_db_person_path(@person), notice: t("resources.person.created")
    end

    def edit
      authorize @person, :edit?
    end

    def update(person)
      authorize @person, :update?

      @person.attributes = person
      @person.user = current_user

      return render(:edit) unless @person.valid?
      @person.save_and_create_activity!

      redirect_to edit_db_person_path(@person), notice: t("resources.person.updated")
    end

    def hide
      authorize @person, :hide?

      @person.hide!

      flash[:notice] = t("resources.person.unpublished")
      redirect_back fallback_location: db_people_path
    end

    def destroy
      @person.destroy

      flash[:notice] = t("resources.person.deleted")
      redirect_back fallback_location: db_people_path
    end

    def activities
      @activities = @person.db_activities.order(id: :desc)
      @comment = @person.db_comments.new
    end

    private

    def load_person
      @person = Person.find(params[:id])
    end
  end
end
