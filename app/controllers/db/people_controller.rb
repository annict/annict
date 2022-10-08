# frozen_string_literal: true

module Db
  class PeopleController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @people = Person
        .without_deleted
        .order(id: :desc)
        .page(params[:page])
        .per(100)
    end

    def new
      @form = Deprecated::Db::PersonRowsForm.new
      authorize @form
    end

    def create
      @form = Deprecated::Db::PersonRowsForm.new(person_rows_form_params)
      @form.user = current_user
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_person_list_path, notice: t("resources.person.created")
    end

    def edit
      @person = Person.without_deleted.find(params[:id])
      authorize @person
    end

    def update
      @person = Person.without_deleted.find(params[:id])
      authorize @person

      @person.attributes = person_params
      @person.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @person.valid?

      @person.save_and_create_activity!

      redirect_to db_edit_person_path(@person), notice: t("resources.person.updated")
    end

    def destroy
      @person = Person.without_deleted.find(params[:id])
      authorize @person

      @person.destroy_in_batches

      redirect_back(
        fallback_location: db_person_list_path,
        notice: t("messages._common.deleted")
      )
    end

    private

    def person_rows_form_params
      params.require(:db_person_rows_form).permit(:rows)
    end

    def person_params
      params.require(:person).permit(
        :name, :name_kana, :name_en, :nickname, :nickname_en, :gender,
        :blood_type, :prefecture_id, :birthday, :height, :url, :url_en,
        :wikipedia_url, :wikipedia_url_en, :twitter_username, :twitter_username_en
      )
    end
  end
end
