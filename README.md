# users_rest_api

Сервис, предоставляющий API для работы с данными пользователей.

## Пример пользователя (JSON)

```
{
    "surname": "Иванов",            // обязательное поле
    "name": "Иван",                 // обязательное поле
    "patronymic": "Иванович",       // может быть пустым
    "sex": "М",                     // обязательное поле
    "status": "active",             // обязательное поле
    "date_of_birth": "1995-01-01",  // может быть пустым
    "date_added": "2024-01-31",     // заполняется автоматически
    "id": 1                         // заполняется автоматически
}
```

## Доступные методы API:

- Создание пользователя

  **POST** **/users/new/json**

  Пример запроса:
  ```
  curl --location 'localhost:8080/users/new/json' \
  --header 'Content-Type: application/json' \
  --data '{
      "surname": "Иванов",
      "name": "Иван",
      "patronymic": "Иванович",
      "sex": "М",
      "status": "active",
      "date_of_birth": "2001-01-01"
  }'
  ```

- Создание пользователей из файла формата XLS

  **POST** **/users/new/xls**

  Пример запроса:
  ```
  curl --location 'localhost:8080/users/new/xls' --form 'file=@"/C:/Users/User1/Desktop/users.xls"'
  ```

  API поддерживает файлы с несколькими листами. Каждый лист должен содержать 5 колонок: Фамилия, Имя, Отчество, Пол, Почта

- Создание пользователей из файла формата XLSX

  **POST** **/users/new/xlsx**

  Пример запроса:
  ```
  curl --location 'localhost:8080/users/new/xlsx' --form 'file=@"/C:/Users/User1/Desktop/users.xlsx"'
  ```

  API поддерживает файлы с несколькими листами. Каждый лист должен содержать 5 колонок: Фамилия, Имя, Отчество, Пол, Почта

- Редактирование пользователя

  **PUT** **/users/{id}**

  Пример запроса:
  ```
  curl --location --request PUT 'localhost:8080/users/1' \
  --header 'Content-Type: application/json' \
  --data '{
      "surname": "Иванов",
      "name": "Иван",
      "patronymic": "Иванович",
      "sex": "М",
      "status": "active",
      "date_of_birth": "2001-01-01"
  }'
  ```

- Удаление пользователя

  **DELETE** **/users/{id}**

  Пример запроса:
  ```
  curl --location --request DELETE 'localhost:8080/users/1'
  ```

- Получение пользователя по id

  **GET** **/users/{id}**

  Пример запроса:
  ```
  curl --location 'localhost:8080/users/1'
  ```

- Поиск пользователей

  **GET** **/users**

  Возможные параметры:
  - filter (позволяет фильтровать список пользователей по значению атрибутов)
    
      Возможные запросы:
      -  По полу:
        
          -  **/users?filter=sex."М"**
          -  **/users?filter=sex."Ж"**
      
      -  По статусу:
      
          -  **/users?filter=status."active"**
          -  **/users?filter=status."banned"**
          -  **/users?filter=status."deleted"**
      
      -  По полному имени:
    
          -  **/users?filter=fullname."Иванов Иван Иванович"**
    
  - sortBy (задает параметры сортировки; имеет форму **имя_поля.направление**, где **направление** либо *asc* (по возрастанию), либо *desc* (по убыванию))

      Пример: **/users?filter=fullname."Иванов Иван Иванович"&sortBy=surname.asc**
      
  - limit (ограничивает количество возвращаемых пользователей; задается ненулевым положительным числом)
 
      Пример: **/users?limit=10**
     
  - offset (устанавливает смещение по списку; задается ненулевым положительным числом)
    
      Пример: **/users?limit=10&offset=5**
    
