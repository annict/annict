module Db
  class DraftPeopleController < Db::ApplicationController
    permits :name, :name_kana, :nickname, :gender, :blood_type, :prefecture_id,
            :birthday, :height, :url, :wikipedia_url, :twitter_username,
            :person_id, edit_request_attributes: [:id, :title, :body]

    before_action :authenticate_user!

    def new(person_id: nil)
      @draft_person = if person_id.present?
        @person = Person.find(person_id)
        DraftPerson.new(@person.attributes.slice(*Person::DIFF_FIELDS.map(&:to_s)))
      else
        DraftPerson.new
      end
      @draft_person.build_edit_request
    end

    def create(draft_person)
      binding.pry
      @draft_person = DraftPerson.new(draft_person)
      @draft_person.edit_request.user = current_user

      if draft_person[:person_id].present?
        @person = Person.find(draft_person[:person_id])
        @draft_person.origin = @person
      end

      if @draft_person.save
        flash[:notice] = "編集リクエストを作成しました"
        redirect_to db_edit_request_path(@draft_person.edit_request)
      else
        render :new
      end
    end

    def edit(id)
      @draft_person = DraftPerson.find(id)
      authorize @draft_person, :edit?
    end

    def update(id, draft_person)
      @draft_person = DraftPerson.find(id)
      authorize @draft_person, :update?

      if @draft_person.update(draft_person)
        flash[:notice] = "作品の編集リクエストを更新しました"
        redirect_to db_edit_request_path(@draft_person.edit_request)
      else
        render :edit
      end
    end
  end
end
