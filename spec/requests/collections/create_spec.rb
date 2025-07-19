# typed: false
# frozen_string_literal: true

RSpec.describe "POST /collections", type: :request do
  it "ログインしているとき、正常なパラメーターでコレクションが作成されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/collections", params: {
        forms_collection_form: {
          name: "テストコレクション",
          description: "テストの説明"
        }
      }
    }.to change(Collection, :count).by(1)

    collection = Collection.last
    expect(collection.name).to eq("テストコレクション")
    expect(collection.description).to eq("テストの説明")
    expect(collection.user).to eq(user)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to be_present
  end

  it "ログインしているとき、説明なしでコレクションが作成されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/collections", params: {
        forms_collection_form: {
          name: "テストコレクション",
          description: ""
        }
      }
    }.to change(Collection, :count).by(1)

    collection = Collection.last
    expect(collection.name).to eq("テストコレクション")
    expect(collection.description).to eq("")
    expect(collection.user).to eq(user)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
  end
end
