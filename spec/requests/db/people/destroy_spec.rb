# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/people/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    person = create(:person, :not_deleted)

    expect(Person.count).to eq(1)

    delete "/db/people/#{person.id}"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(Person.count).to eq(1)
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    person = create(:person, :not_deleted)
    login_as(user, scope: :user)

    expect(Person.count).to eq(1)

    delete "/db/people/#{person.id}"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Person.count).to eq(1)
  end

  it "編集者権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    person = create(:person, :not_deleted)
    login_as(user, scope: :user)

    expect(Person.count).to eq(1)

    delete "/db/people/#{person.id}"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(Person.count).to eq(1)
  end

  it "管理者権限を持つユーザーがログインしているとき、人物を論理削除できること" do
    user = create(:registered_user, :with_admin_role)
    person = create(:person, :not_deleted)
    login_as(user, scope: :user)

    expect(Person.count).to eq(1)

    delete "/db/people/#{person.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")

    expect(Person.count).to eq(0)
  end

  it "存在しない人物IDを指定したとき、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect { delete "/db/people/invalid-id" }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みの人物を削除しようとしたとき、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user, :with_admin_role)
    person = create(:person, :not_deleted)
    person.destroy_in_batches
    login_as(user, scope: :user)

    expect { delete "/db/people/#{person.id}" }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
