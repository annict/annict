# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/people", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    person_params = {
      rows: "徳川家康,とくがわいえやす"
    }

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Person.all.size).to eq(0)
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスが拒否されること" do
    user = create(:registered_user)
    person_params = {
      rows: "徳川家康,とくがわいえやす"
    }

    login_as(user, scope: :user)

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Person.all.size).to eq(0)
  end

  it "編集者権限を持つユーザーがログインしているとき、人物を作成できること" do
    user = create(:registered_user, :with_editor_role)
    person_params = {
      rows: "徳川家康,とくがわいえやす"
    }

    login_as(user, scope: :user)

    expect(Person.all.size).to eq(0)

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Person.all.size).to eq(1)

    person = Person.first
    expect(person.name).to eq("徳川家康")
    expect(person.name_kana).to eq("とくがわいえやす")
  end

  it "編集者権限を持つユーザーがログインしているとき、複数の人物を一度に作成できること" do
    user = create(:registered_user, :with_editor_role)
    person_params = {
      rows: "徳川家康,とくがわいえやす\n織田信長,おだのぶなが\n豊臣秀吉,とよとみひでよし"
    }

    login_as(user, scope: :user)

    expect(Person.all.size).to eq(0)

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Person.all.size).to eq(3)

    people = Person.order(:created_at)
    expect(people[0].name).to eq("徳川家康")
    expect(people[0].name_kana).to eq("とくがわいえやす")
    expect(people[1].name).to eq("織田信長")
    expect(people[1].name_kana).to eq("おだのぶなが")
    expect(people[2].name).to eq("豊臣秀吉")
    expect(people[2].name_kana).to eq("とよとみひでよし")
  end

  it "編集者権限を持つユーザーがログインしているとき、名前のみ（ふりがななし）で人物を作成できること" do
    user = create(:registered_user, :with_editor_role)
    person_params = {
      rows: "徳川家康,"
    }

    login_as(user, scope: :user)

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Person.all.size).to eq(1)

    person = Person.first
    expect(person.name).to eq("徳川家康")
    expect(person.name_kana).to eq("")
  end

  it "編集者権限を持つユーザーがログインしているとき、入力が空の場合はバリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    person_params = {
      rows: ""
    }

    login_as(user, scope: :user)

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(422)
    expect(Person.all.size).to eq(0)
  end

  it "編集者権限を持つユーザーがログインしているとき、重複する名前の場合はバリデーションエラーになること" do
    user = create(:registered_user, :with_editor_role)
    create(:person, name: "徳川家康")
    person_params = {
      rows: "徳川家康,とくがわいえやす"
    }

    login_as(user, scope: :user)

    post "/db/people", params: {deprecated_db_person_rows_form: person_params}

    expect(response.status).to eq(422)
    expect(Person.all.size).to eq(1)
  end
end
