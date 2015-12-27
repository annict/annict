module Db
  class PeopleController < Db::ApplicationController
    permits :name, :name_kana, :nickname, :gender, :blood_type, :prefecture_id,
            :birthday, :height, :url, :wikipedia_url, :twitter_username

    before_action :authenticate_user!, only: [:new, :create, :edit, :update, :destroy]

    def index(page: nil)
      @people = Person.order(id: :desc).page(page)
    end

    def new
      @person = Person.new
      authorize @person, :new?
    end

    def create(person)
      @person = Person.new(person)
      authorize @person, :create?

      if @person.save_and_create_db_activity(current_user, "people.create")
        redirect_to edit_db_person_path(@person), notice: "登録しました"
      else
        render :new
      end
    end

    def edit(id)
      @person = Person.find(id)
      authorize @person, :edit?
    end

    def update(id, person)
      @person = Person.find(id)
      authorize @person, :update?

      @person.attributes = person
      if @person.save_and_create_db_activity(current_user, "people.update")
        redirect_to edit_db_person_path(@person), notice: "更新しました"
      else
        render :edit
      end
    end
  end
end
