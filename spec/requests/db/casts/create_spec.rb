# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/casts", type: :request do
  it "ログインしていない場合、ログインページにリダイレクトすること" do
    character = create(:character)
    person = create(:person)
    work = create(:work)
    form_params = {
      rows: "#{character.id},#{person.id}"
    }

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Cast.all.size).to eq(0)
  end

  it "編集者権限がないユーザーの場合、アクセスできないこと" do
    character = create(:character)
    person = create(:person)
    work = create(:work)
    user = create(:registered_user)
    form_params = {
      rows: "#{character.id},#{person.id}"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Cast.all.size).to eq(0)
  end

  it "編集者権限を持つユーザーの場合、キャストを作成できること" do
    character = create(:character)
    person = create(:person)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{character.id},#{person.id}"
    }

    login_as(user, scope: :user)

    expect(Cast.all.size).to eq(0)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Cast.all.size).to eq(1)

    cast = Cast.last
    expect(cast.character_id).to eq(character.id)
    expect(cast.person_id).to eq(person.id)
  end

  it "編集者権限を持つユーザーの場合、複数のキャストを一度に作成できること" do
    character1 = create(:character)
    person1 = create(:person)
    character2 = create(:character)
    person2 = create(:person)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{character1.id},#{person1.id}\n#{character2.id},#{person2.id}"
    }

    login_as(user, scope: :user)

    expect(Cast.all.size).to eq(0)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Cast.all.size).to eq(2)

    casts = Cast.order(:sort_number)
    expect(casts[0].character_id).to eq(character1.id)
    expect(casts[0].person_id).to eq(person1.id)
    expect(casts[0].sort_number).to eq(0)
    expect(casts[1].character_id).to eq(character2.id)
    expect(casts[1].person_id).to eq(person2.id)
    expect(casts[1].sort_number).to eq(10)
  end

  it "編集者権限を持つユーザーの場合、キャラクター名と人物名でキャストを作成できること" do
    character = create(:character, name: "綾波レイ")
    person = create(:person, name: "林原めぐみ")
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "綾波レイ,林原めぐみ"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Cast.all.size).to eq(1)

    cast = Cast.last
    expect(cast.character_id).to eq(character.id)
    expect(cast.person_id).to eq(person.id)
  end

  it "存在しないキャラクターIDが指定された場合、バリデーションエラーになること" do
    person = create(:person)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "999999,#{person.id}"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(Cast.all.size).to eq(0)
  end

  it "存在しない人物IDが指定された場合、バリデーションエラーになること" do
    character = create(:character)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{character.id},999999"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(Cast.all.size).to eq(0)
  end

  it "空のrowsパラメータが送られた場合、バリデーションエラーになること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: ""
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/casts", params: {deprecated_db_cast_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(Cast.all.size).to eq(0)
  end
end
