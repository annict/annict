# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /collections/:collection_id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)

    delete collection_path(collection.id)

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき自分のコレクションを削除すると、コレクション一覧ページにリダイレクトされること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    login_as(user, scope: :user)

    delete collection_path(collection.id)

    expect(response).to redirect_to(user_collection_list_path(user.username))
    expect(flash[:notice]).to eq(I18n.t("messages._common.deleted"))
  end

  it "ログインしているとき自分のコレクションを削除すると、コレクションが物理削除されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)
    collection_id = collection.id
    login_as(user, scope: :user)

    delete collection_path(collection.id)

    expect {
      Collection.find(collection_id)
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき他人のコレクションを削除しようとすると、404エラーが返されること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    collection = create(:collection, user: other_user)
    login_as(user, scope: :user)

    expect {
      delete collection_path(collection.id)
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき存在しないコレクションを削除しようとすると、404エラーが返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      delete collection_path("nonexistent")
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき削除済みのコレクションを削除しようとすると、404エラーが返されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      delete collection_path(collection.id)
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
